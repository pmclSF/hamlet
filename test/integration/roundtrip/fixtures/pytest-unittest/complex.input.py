import pytest

from app.services.notification import NotificationService, NotificationError
from app.clients.email import EmailClient
from app.clients.sms import SmsClient


@pytest.fixture
def email_client(monkeypatch):
    client = EmailClient(api_key='test-key')
    sent = []
    monkeypatch.setattr(client, 'send', lambda to, subject, body: sent.append({
        'to': to,
        'subject': subject,
        'body': body,
    }))
    client._sent = sent
    return client


@pytest.fixture
def sms_client(monkeypatch):
    client = SmsClient(api_key='test-key')
    sent = []
    monkeypatch.setattr(client, 'send', lambda to, message: sent.append({
        'to': to,
        'message': message,
    }))
    client._sent = sent
    return client


@pytest.fixture
def notification_service(email_client, sms_client):
    return NotificationService(email_client=email_client, sms_client=sms_client)


def test_sends_email_notification(notification_service, email_client):
    notification_service.notify(
        channel='email',
        recipient='user@example.com',
        subject='Welcome',
        body='Thanks for signing up.',
    )
    assert len(email_client._sent) == 1
    assert email_client._sent[0]['to'] == 'user@example.com'


def test_sends_sms_notification(notification_service, sms_client):
    notification_service.notify(
        channel='sms',
        recipient='+15551234567',
        body='Your code is 1234.',
    )
    assert len(sms_client._sent) == 1
    assert '1234' in sms_client._sent[0]['message']


def test_raises_for_unknown_channel(notification_service):
    with pytest.raises(NotificationError, match='Unsupported channel: pigeon'):
        notification_service.notify(
            channel='pigeon',
            recipient='bird@example.com',
            body='coo',
        )


def test_sends_to_multiple_recipients(notification_service, email_client):
    recipients = ['a@example.com', 'b@example.com', 'c@example.com']
    for r in recipients:
        notification_service.notify(
            channel='email',
            recipient=r,
            subject='Broadcast',
            body='Important update.',
        )
    assert len(email_client._sent) == 3


@pytest.mark.parametrize('channel, recipient', [
    ('email', 'admin@example.com'),
    ('sms', '+15559876543'),
])
def test_notify_returns_confirmation(notification_service, channel, recipient):
    result = notification_service.notify(
        channel=channel,
        recipient=recipient,
        subject='Test',
        body='Hello.',
    )
    assert result['status'] == 'sent'
    assert result['channel'] == channel


@pytest.fixture
def failing_email_client(monkeypatch):
    client = EmailClient(api_key='test-key')
    monkeypatch.setattr(client, 'send', lambda *args: (_ for _ in ()).throw(
        ConnectionError('SMTP server unreachable')
    ))
    return client


@pytest.fixture
def unreliable_service(failing_email_client, sms_client):
    return NotificationService(email_client=failing_email_client, sms_client=sms_client)


def test_wraps_transport_errors(unreliable_service):
    with pytest.raises(NotificationError, match='Failed to send via email'):
        unreliable_service.notify(
            channel='email',
            recipient='user@example.com',
            subject='Oops',
            body='This will fail.',
        )


def test_sms_still_works_when_email_fails(unreliable_service, sms_client):
    unreliable_service.notify(
        channel='sms',
        recipient='+15551112222',
        body='Fallback message.',
    )
    assert len(sms_client._sent) == 1
    assert sms_client._sent[0]['to'] == '+15551112222'

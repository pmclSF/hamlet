# pytest test for Django REST API views
# Inspired by real-world pytest-django view tests for a task management API

import pytest
import json
from datetime import datetime, timedelta
from django.test import RequestFactory
from django.contrib.auth.models import User

from tasks.models import Task, Project
from tasks.views import TaskViewSet


@pytest.fixture
def request_factory():
    return RequestFactory()


@pytest.fixture
def admin_user(db):
    user = User.objects.create_user(
        username='admin',
        email='admin@example.com',
        password='testpass123',
        is_staff=True,
    )
    yield user
    user.delete()


@pytest.fixture
def project(db, admin_user):
    proj = Project.objects.create(
        name='Backend Refactor',
        owner=admin_user,
        deadline=datetime.now() + timedelta(days=30),
    )
    yield proj
    proj.delete()


@pytest.fixture
def sample_tasks(db, project, admin_user):
    tasks = [
        Task.objects.create(
            title='Set up CI pipeline',
            project=project,
            assignee=admin_user,
            status='done',
            priority=1,
        ),
        Task.objects.create(
            title='Write integration tests',
            project=project,
            assignee=admin_user,
            status='in_progress',
            priority=2,
        ),
        Task.objects.create(
            title='Deploy to staging',
            project=project,
            assignee=admin_user,
            status='todo',
            priority=3,
        ),
    ]
    yield tasks
    Task.objects.filter(project=project).delete()


def test_list_tasks_returns_all_tasks_for_project(request_factory, admin_user, sample_tasks, project):
    request = request_factory.get(f'/api/projects/{project.id}/tasks/')
    request.user = admin_user

    view = TaskViewSet.as_view({'get': 'list'})
    response = view(request, project_id=project.id)

    assert response.status_code == 200
    assert len(response.data) == 3


def test_list_tasks_filters_by_status(request_factory, admin_user, sample_tasks, project):
    request = request_factory.get(f'/api/projects/{project.id}/tasks/', {'status': 'todo'})
    request.user = admin_user

    view = TaskViewSet.as_view({'get': 'list'})
    response = view(request, project_id=project.id)

    assert response.status_code == 200
    assert len(response.data) == 1
    assert response.data[0]['title'] == 'Deploy to staging'


def test_create_task_with_valid_data(request_factory, admin_user, project):
    payload = {
        'title': 'Review pull requests',
        'status': 'todo',
        'priority': 1,
    }
    request = request_factory.post(
        f'/api/projects/{project.id}/tasks/',
        data=json.dumps(payload),
        content_type='application/json',
    )
    request.user = admin_user

    view = TaskViewSet.as_view({'post': 'create'})
    response = view(request, project_id=project.id)

    assert response.status_code == 201
    assert response.data['title'] == 'Review pull requests'
    assert 'id' in response.data


def test_create_task_without_title_returns_400(request_factory, admin_user, project):
    payload = {'status': 'todo', 'priority': 1}
    request = request_factory.post(
        f'/api/projects/{project.id}/tasks/',
        data=json.dumps(payload),
        content_type='application/json',
    )
    request.user = admin_user

    view = TaskViewSet.as_view({'post': 'create'})
    response = view(request, project_id=project.id)

    assert response.status_code == 400
    assert 'title' in response.data


@pytest.mark.parametrize('invalid_status', ['invalid', 'completed', '', 'DONE'])
def test_create_task_with_invalid_status_returns_400(request_factory, admin_user, project, invalid_status):
    payload = {'title': 'Bad status task', 'status': invalid_status, 'priority': 1}
    request = request_factory.post(
        f'/api/projects/{project.id}/tasks/',
        data=json.dumps(payload),
        content_type='application/json',
    )
    request.user = admin_user

    view = TaskViewSet.as_view({'post': 'create'})
    response = view(request, project_id=project.id)

    assert response.status_code == 400


def test_delete_task_removes_from_database(request_factory, admin_user, sample_tasks):
    task = sample_tasks[0]
    request = request_factory.delete(f'/api/tasks/{task.id}/')
    request.user = admin_user

    view = TaskViewSet.as_view({'delete': 'destroy'})
    response = view(request, pk=task.id)

    assert response.status_code == 204
    with pytest.raises(Task.DoesNotExist):
        Task.objects.get(pk=task.id)


def test_unauthenticated_request_returns_403(request_factory, project):
    from django.contrib.auth.models import AnonymousUser

    request = request_factory.get(f'/api/projects/{project.id}/tasks/')
    request.user = AnonymousUser()

    view = TaskViewSet.as_view({'get': 'list'})
    response = view(request, project_id=project.id)

    assert response.status_code == 403

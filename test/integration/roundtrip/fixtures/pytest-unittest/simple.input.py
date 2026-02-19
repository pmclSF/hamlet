from app.services.slug_generator import generate_slug


def test_generates_slug_from_simple_title():
    result = generate_slug('Hello World')
    assert result == 'hello-world'


def test_strips_special_characters():
    result = generate_slug('Price: $100 & Up!')
    assert result == 'price-100-up'


def test_collapses_consecutive_hyphens():
    result = generate_slug('too   many   spaces')
    assert result == 'too-many-spaces'

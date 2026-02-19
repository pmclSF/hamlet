# unittest test for a standard library string utilities module
# Inspired by real-world Python stdlib-style unit tests

import unittest
from utils.text_processor import TextProcessor


class TestTextProcessor(unittest.TestCase):

    def setUp(self):
        self.processor = TextProcessor()

    def tearDown(self):
        self.processor = None

    def test_slugify_converts_spaces_to_hyphens(self):
        result = self.processor.slugify('Hello World')
        self.assertEqual(result, 'hello-world')

    def test_slugify_removes_special_characters(self):
        result = self.processor.slugify('Hello, World! @2025')
        self.assertEqual(result, 'hello-world-2025')

    def test_slugify_collapses_multiple_hyphens(self):
        result = self.processor.slugify('too   many   spaces')
        self.assertEqual(result, 'too-many-spaces')

    def test_slugify_strips_leading_and_trailing_hyphens(self):
        result = self.processor.slugify('  padded  ')
        self.assertEqual(result, 'padded')

    def test_truncate_shortens_text_to_max_length(self):
        text = 'This is a fairly long sentence that needs to be truncated.'
        result = self.processor.truncate(text, max_length=20)
        self.assertTrue(len(result) <= 20)
        self.assertTrue(result.endswith('...'))

    def test_truncate_does_not_alter_short_text(self):
        text = 'Short'
        result = self.processor.truncate(text, max_length=100)
        self.assertEqual(result, 'Short')
        self.assertNotIn('...', result)

    def test_truncate_raises_on_negative_length(self):
        with self.assertRaises(ValueError) as ctx:
            self.processor.truncate('Hello', max_length=-1)
        self.assertIn('max_length must be positive', str(ctx.exception))

    def test_word_count_returns_correct_count(self):
        result = self.processor.word_count('The quick brown fox jumps')
        self.assertEqual(result, 5)

    def test_word_count_handles_empty_string(self):
        result = self.processor.word_count('')
        self.assertEqual(result, 0)

    def test_word_count_handles_multiple_whitespace(self):
        result = self.processor.word_count('  spaced   out  words  ')
        self.assertEqual(result, 3)

    def test_extract_emails_finds_all_addresses(self):
        text = 'Contact alice@example.com or bob@example.org for info.'
        result = self.processor.extract_emails(text)
        self.assertEqual(len(result), 2)
        self.assertIn('alice@example.com', result)
        self.assertIn('bob@example.org', result)

    def test_extract_emails_returns_empty_list_when_none_found(self):
        result = self.processor.extract_emails('No emails here.')
        self.assertEqual(result, [])

    def test_capitalize_sentences_with_various_inputs(self):
        test_cases = [
            ('hello. world.', 'Hello. World.'),
            ('already Fine. also Good.', 'Already Fine. Also Good.'),
            ('single', 'Single'),
        ]
        for input_text, expected in test_cases:
            with self.subTest(input_text=input_text):
                result = self.processor.capitalize_sentences(input_text)
                self.assertEqual(result, expected)

    def test_reverse_words_flips_word_order(self):
        result = self.processor.reverse_words('one two three')
        self.assertEqual(result, 'three two one')

    def test_is_palindrome_returns_true_for_palindromes(self):
        self.assertTrue(self.processor.is_palindrome('racecar'))
        self.assertTrue(self.processor.is_palindrome('A man a plan a canal Panama'))

    def test_is_palindrome_returns_false_for_non_palindromes(self):
        self.assertFalse(self.processor.is_palindrome('hello'))


if __name__ == '__main__':
    unittest.main()

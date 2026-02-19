// JUnit 5 test for a repository layer with nested tests and parameterized inputs
// Inspired by real-world Spring Data JPA repository tests

package com.example.blog.repository;

import org.junit.jupiter.api.*;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;
import org.junit.jupiter.params.provider.NullAndEmptySource;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;

@DisplayName("ArticleRepository")
public class ArticleRepositoryTest {

    private ArticleRepository repository;

    @BeforeEach
    void setUp() {
        repository = new InMemoryArticleRepository();
        repository.save(new Article("Spring Boot Guide", "intro-to-spring", "Alice", LocalDateTime.of(2025, 1, 10, 9, 0)));
        repository.save(new Article("Testing Best Practices", "testing-best-practices", "Bob", LocalDateTime.of(2025, 2, 15, 14, 30)));
        repository.save(new Article("Docker for Beginners", "docker-beginners", "Alice", LocalDateTime.of(2025, 3, 20, 11, 0)));
    }

    @AfterEach
    void tearDown() {
        repository.deleteAll();
    }

    @Test
    @DisplayName("should return all articles sorted by date descending")
    void findAll_returnsSortedByDateDesc() {
        List<Article> articles = repository.findAllSortedByDate();

        assertEquals(3, articles.size());
        assertEquals("docker-beginners", articles.get(0).getSlug());
        assertEquals("testing-best-practices", articles.get(1).getSlug());
    }

    @Nested
    @DisplayName("findBySlug")
    class FindBySlug {

        @Test
        @DisplayName("should return the article when slug exists")
        void returnsArticle_whenSlugExists() {
            Optional<Article> result = repository.findBySlug("intro-to-spring");

            assertTrue(result.isPresent());
            assertEquals("Spring Boot Guide", result.get().getTitle());
        }

        @Test
        @DisplayName("should return empty when slug does not exist")
        void returnsEmpty_whenSlugNotFound() {
            Optional<Article> result = repository.findBySlug("nonexistent-slug");

            assertFalse(result.isPresent());
        }
    }

    @Nested
    @DisplayName("findByAuthor")
    class FindByAuthor {

        @Test
        @DisplayName("should return all articles by the given author")
        void returnsArticles_forKnownAuthor() {
            List<Article> articles = repository.findByAuthor("Alice");

            assertEquals(2, articles.size());
            assertTrue(articles.stream().allMatch(a -> "Alice".equals(a.getAuthor())));
        }

        @Test
        @DisplayName("should return an empty list for an unknown author")
        void returnsEmpty_forUnknownAuthor() {
            List<Article> articles = repository.findByAuthor("Unknown");

            assertTrue(articles.isEmpty());
        }
    }

    @Nested
    @DisplayName("save validation")
    class SaveValidation {

        @ParameterizedTest
        @NullAndEmptySource
        @DisplayName("should reject articles with null or empty titles")
        void rejectsInvalidTitle(String title) {
            assertThrows(IllegalArgumentException.class, () -> {
                repository.save(new Article(title, "slug", "Author", LocalDateTime.now()));
            });
        }

        @ParameterizedTest
        @CsvSource({
            "'Title One', 'title-one'",
            "'Hello World', 'hello-world'",
            "'JUnit 5 Rocks!', 'junit-5-rocks'"
        })
        @DisplayName("should generate the expected slug from the title")
        void generatesSlug(String title, String expectedSlug) {
            Article article = new Article(title, null, "Author", LocalDateTime.now());
            Article saved = repository.save(article);

            assertEquals(expectedSlug, saved.getSlug());
        }

        @Test
        @DisplayName("should set id and timestamps on save")
        void setsIdAndTimestamps() {
            Article article = new Article("New Article", "new-article", "Carol", null);
            Article saved = repository.save(article);

            assertAll(
                () -> assertNotNull(saved.getId()),
                () -> assertNotNull(saved.getCreatedAt()),
                () -> assertNotNull(saved.getUpdatedAt())
            );
        }
    }
}

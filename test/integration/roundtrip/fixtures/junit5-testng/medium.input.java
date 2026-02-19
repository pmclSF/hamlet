package com.example.service;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

@DisplayName("UserService")
class UserServiceTest {

    private UserService userService;

    @BeforeEach
    void setUp() {
        userService = new UserService(new InMemoryUserRepository());
    }

    @Test
    @DisplayName("should create a new user with a generated ID")
    void shouldCreateUser() {
        User user = userService.createUser("alice@example.com", "Alice");
        assertNotNull(user.getId());
        assertEquals("alice@example.com", user.getEmail());
    }

    @Test
    @DisplayName("should throw when creating a user with a duplicate email")
    void shouldRejectDuplicateEmail() {
        userService.createUser("alice@example.com", "Alice");
        assertThrows(DuplicateEmailException.class, () ->
            userService.createUser("alice@example.com", "Alice 2")
        );
    }

    @Nested
    @DisplayName("when finding users")
    class FindUsers {

        @BeforeEach
        void createSampleUsers() {
            userService.createUser("alice@example.com", "Alice");
            userService.createUser("bob@example.com", "Bob");
        }

        @Test
        @DisplayName("should find a user by email")
        void shouldFindByEmail() {
            User found = userService.findByEmail("bob@example.com");
            assertEquals("Bob", found.getName());
        }

        @Test
        @DisplayName("should return null for unknown email")
        void shouldReturnNullForUnknown() {
            User found = userService.findByEmail("nobody@example.com");
            assertEquals(null, found);
        }

        @ParameterizedTest
        @CsvSource({
            "alice@example.com, Alice",
            "bob@example.com, Bob"
        })
        @DisplayName("should resolve correct name by email")
        void shouldResolveNameByEmail(String email, String expectedName) {
            User found = userService.findByEmail(email);
            assertEquals(expectedName, found.getName());
        }
    }
}

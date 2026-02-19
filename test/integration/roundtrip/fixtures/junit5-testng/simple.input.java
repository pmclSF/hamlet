package com.example.service;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

class StringCalculatorTest {

    private StringCalculator calculator;

    @BeforeEach
    void setUp() {
        calculator = new StringCalculator();
    }

    @Test
    void shouldReturnZeroForEmptyString() {
        int result = calculator.add("");
        assertEquals(0, result);
    }

    @Test
    void shouldReturnNumberForSingleValue() {
        int result = calculator.add("5");
        assertEquals(5, result);
    }

    @Test
    void shouldSumTwoNumbers() {
        int result = calculator.add("3,7");
        assertEquals(10, result);
    }
}

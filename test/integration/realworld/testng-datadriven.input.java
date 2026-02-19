// TestNG data-driven test for a currency conversion service
// Inspired by real-world TestNG tests for financial calculations

package com.example.finance.service;

import org.testng.annotations.BeforeMethod;
import org.testng.annotations.AfterMethod;
import org.testng.annotations.DataProvider;
import org.testng.annotations.Test;
import org.testng.Assert;

import java.math.BigDecimal;
import java.math.RoundingMode;

public class CurrencyConverterTest {

    private CurrencyConverter converter;
    private ExchangeRateProvider rateProvider;

    @BeforeMethod
    public void setUp() {
        rateProvider = new InMemoryExchangeRateProvider();
        rateProvider.setRate("USD", "EUR", new BigDecimal("0.92"));
        rateProvider.setRate("USD", "GBP", new BigDecimal("0.79"));
        rateProvider.setRate("EUR", "USD", new BigDecimal("1.09"));
        rateProvider.setRate("GBP", "USD", new BigDecimal("1.27"));
        rateProvider.setRate("USD", "JPY", new BigDecimal("149.50"));
        converter = new CurrencyConverter(rateProvider);
    }

    @AfterMethod
    public void tearDown() {
        rateProvider.clearRates();
    }

    @DataProvider(name = "conversionData")
    public Object[][] conversionData() {
        return new Object[][] {
            { "USD", "EUR", "100.00", "92.00" },
            { "USD", "GBP", "100.00", "79.00" },
            { "EUR", "USD", "50.00", "54.50" },
            { "GBP", "USD", "200.00", "254.00" },
            { "USD", "JPY", "10.00", "1495.00" },
        };
    }

    @Test(dataProvider = "conversionData")
    public void convert_withKnownRates_returnsExpectedAmount(
            String from, String to, String amount, String expected) {
        BigDecimal result = converter.convert(
            new BigDecimal(amount), from, to
        );

        Assert.assertEquals(
            result.setScale(2, RoundingMode.HALF_UP),
            new BigDecimal(expected)
        );
    }

    @Test
    public void convert_sameSourceAndTarget_returnsSameAmount() {
        BigDecimal amount = new BigDecimal("250.00");
        BigDecimal result = converter.convert(amount, "USD", "USD");

        Assert.assertEquals(result, amount);
    }

    @Test(expectedExceptions = UnsupportedCurrencyException.class)
    public void convert_withUnsupportedCurrency_throwsException() {
        converter.convert(new BigDecimal("100.00"), "USD", "XYZ");
    }

    @Test
    public void convert_withZeroAmount_returnsZero() {
        BigDecimal result = converter.convert(BigDecimal.ZERO, "USD", "EUR");

        Assert.assertEquals(result.compareTo(BigDecimal.ZERO), 0);
    }

    @Test(expectedExceptions = IllegalArgumentException.class)
    public void convert_withNegativeAmount_throwsException() {
        converter.convert(new BigDecimal("-50.00"), "USD", "EUR");
    }

    @DataProvider(name = "roundingData")
    public Object[][] roundingData() {
        return new Object[][] {
            { "USD", "EUR", "33.33", "30.66" },
            { "USD", "GBP", "77.77", "61.44" },
        };
    }

    @Test(dataProvider = "roundingData")
    public void convert_appliesCorrectRounding(
            String from, String to, String amount, String expected) {
        BigDecimal result = converter.convert(
            new BigDecimal(amount), from, to
        );

        Assert.assertEquals(
            result.setScale(2, RoundingMode.HALF_UP),
            new BigDecimal(expected)
        );
    }

    @Test
    public void getAvailableCurrencies_returnsAllConfiguredPairs() {
        var currencies = converter.getAvailableCurrencies();

        Assert.assertTrue(currencies.contains("USD"));
        Assert.assertTrue(currencies.contains("EUR"));
        Assert.assertTrue(currencies.contains("GBP"));
        Assert.assertTrue(currencies.contains("JPY"));
        Assert.assertEquals(currencies.size(), 4);
    }
}

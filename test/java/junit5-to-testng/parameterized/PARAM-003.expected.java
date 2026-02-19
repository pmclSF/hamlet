import org.testng.Assert;

public class CsvSourceTest {
    // HAMLET-TODO [UNCONVERTIBLE-PARAMETERIZED-TEST]: JUnit 5 @ParameterizedTest requires manual conversion to TestNG @DataProvider
    // Original: @ParameterizedTest
    // Manual action required: Create a @DataProvider method and reference it with @Test(dataProvider = "...")
    @ParameterizedTest
    // HAMLET-TODO [UNCONVERTIBLE-CSV-SOURCE]: JUnit 5 @CsvSource has no direct TestNG equivalent
    // Original: @CsvSource({"1, 1", "2, 4", "3, 9"})
    // Manual action required: Convert CSV data into a @DataProvider method returning Object[][]
    @CsvSource({"1, 1", "2, 4", "3, 9"})
    public void testSquare(int input, int expected) {
        Assert.assertEquals(input * input, expected);
    }
}

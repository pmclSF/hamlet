import org.testng.Assert;

public class ValueSourceTest {
    // TERRAIN-TODO [UNCONVERTIBLE-PARAMETERIZED-TEST]: JUnit 5 @ParameterizedTest requires manual conversion to TestNG @DataProvider
    // Original: @ParameterizedTest
    // Manual action required: Create a @DataProvider method and reference it with @Test(dataProvider = "...")
    @ParameterizedTest
    // TERRAIN-TODO [UNCONVERTIBLE-VALUE-SOURCE]: JUnit 5 @ValueSource has no direct TestNG equivalent
    // Original: @ValueSource(strings = {"hello", "world"})
    // Manual action required: Convert values into a @DataProvider method returning Object[][]
    @ValueSource(strings = {"hello", "world"})
    public void testStrings(String value) {
        Assert.assertNotNull(value);
    }
}

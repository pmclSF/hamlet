import org.testng.Assert;

public class NullEmptyTest {
    // HAMLET-TODO [UNCONVERTIBLE-PARAMETERIZED-TEST]: JUnit 5 @ParameterizedTest requires manual conversion to TestNG @DataProvider
    // Original: @ParameterizedTest
    // Manual action required: Create a @DataProvider method and reference it with @Test(dataProvider = "...")
    @ParameterizedTest
    // HAMLET-TODO [UNCONVERTIBLE-NULL-EMPTY-SOURCE]: JUnit 5 @NullAndEmptySource/@NullSource/@EmptySource has no TestNG equivalent
    // Original: @NullAndEmptySource
    // Manual action required: Add null/empty values to the @DataProvider data set
    @NullAndEmptySource
    public void testNullEmpty(String value) {
        Assert.assertTrue(value == null || value.isEmpty());
    }
}

import org.testng.Assert;
import java.time.Month;

public class EnumSourceTest {
    // HAMLET-TODO [UNCONVERTIBLE-PARAMETERIZED-TEST]: JUnit 5 @ParameterizedTest requires manual conversion to TestNG @DataProvider
    // Original: @ParameterizedTest
    // Manual action required: Create a @DataProvider method and reference it with @Test(dataProvider = "...")
    @ParameterizedTest
    // HAMLET-TODO [UNCONVERTIBLE-ENUM-SOURCE]: JUnit 5 @EnumSource has no direct TestNG equivalent
    // Original: @EnumSource(Month.class)
    // Manual action required: Create a @DataProvider method that returns enum values
    @EnumSource(Month.class)
    public void testMonth(Month month) {
        Assert.assertNotNull(month);
    }
}

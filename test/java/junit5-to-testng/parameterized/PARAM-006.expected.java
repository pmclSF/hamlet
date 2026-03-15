import org.testng.annotations.Test;
import org.testng.Assert;
import java.util.List;

public class MultiImportTest {
    @Test
    public void testBasic() {
        Assert.assertTrue(true);
    }

    // TERRAIN-TODO [UNCONVERTIBLE-PARAMETERIZED-TEST]: JUnit 5 @ParameterizedTest requires manual conversion to TestNG @DataProvider
    // Original: @ParameterizedTest
    // Manual action required: Create a @DataProvider method and reference it with @Test(dataProvider = "...")
    @ParameterizedTest
    // TERRAIN-TODO [UNCONVERTIBLE-VALUE-SOURCE]: JUnit 5 @ValueSource has no direct TestNG equivalent
    // Original: @ValueSource(ints = {1, 2, 3})
    // Manual action required: Convert values into a @DataProvider method returning Object[][]
    @ValueSource(ints = {1, 2, 3})
    public void testInts(int value) {
        Assert.assertTrue(value > 0);
    }
}

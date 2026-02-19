import org.testng.Assert;
import java.util.stream.Stream;

public class MethodSourceTest {
    static Stream<String> stringProvider() {
        return Stream.of("apple", "banana");
    }

    // HAMLET-TODO [UNCONVERTIBLE-PARAMETERIZED-TEST]: JUnit 5 @ParameterizedTest requires manual conversion to TestNG @DataProvider
    // Original: @ParameterizedTest
    // Manual action required: Create a @DataProvider method and reference it with @Test(dataProvider = "...")
    @ParameterizedTest
    // HAMLET-TODO [UNCONVERTIBLE-METHOD-SOURCE]: JUnit 5 @MethodSource should be converted to TestNG @DataProvider
    // Original: @MethodSource("stringProvider")
    // Manual action required: Rename the source method and annotate it with @DataProvider
    @MethodSource("stringProvider")
    public void testFruit(String fruit) {
        Assert.assertNotNull(fruit);
    }
}

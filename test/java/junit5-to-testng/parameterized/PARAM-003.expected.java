import org.testng.Assert;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;

public class CsvSourceTest {
    @ParameterizedTest
    @CsvSource({"1, 1", "2, 4", "3, 9"})
    public void testSquare(int input, int expected) {
        Assert.assertEquals(input * input, expected);
    }
}

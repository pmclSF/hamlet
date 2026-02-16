import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.provider.MethodSource;
import org.junit.jupiter.api.Assertions;

public class DataProviderTest {
    @DataProvider(name = "numbers")
    public Object[][] numberProvider() {
        return new Object[][] {{1, 1}, {2, 4}, {3, 9}};
    }

    @Test(dataProvider = "numbers")
    public void testSquare(int input, int expected) {
        Assertions.assertEquals(expected, input * input);
    }
}

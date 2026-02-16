import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.provider.MethodSource;
import org.junit.jupiter.api.Assertions;

public class MultiConsumerTest {
    @DataProvider(name = "items")
    public Object[][] itemProvider() {
        return new Object[][] {{"apple", 1}, {"banana", 2}};
    }

    @Test(dataProvider = "items")
    public void testFirst(String name, int count) {
        Assertions.assertNotNull(name);
    }

    @Test(dataProvider = "items")
    public void testSecond(String name, int count) {
        Assertions.assertTrue(count > 0);
    }
}

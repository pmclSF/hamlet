import org.testng.annotations.Test;
import org.testng.annotations.DataProvider;
import org.testng.Assert;

public class MultiConsumerTest {
    @DataProvider(name = "items")
    public Object[][] itemProvider() {
        return new Object[][] {{"apple", 1}, {"banana", 2}};
    }

    @Test(dataProvider = "items")
    public void testFirst(String name, int count) {
        Assert.assertNotNull(name);
    }

    @Test(dataProvider = "items")
    public void testSecond(String name, int count) {
        Assert.assertTrue(count > 0);
    }
}

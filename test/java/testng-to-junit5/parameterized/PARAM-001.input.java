import org.testng.annotations.Test;
import org.testng.annotations.DataProvider;
import org.testng.Assert;

public class DataProviderTest {
    @DataProvider(name = "numbers")
    public Object[][] numberProvider() {
        return new Object[][] {{1, 1}, {2, 4}, {3, 9}};
    }

    @Test(dataProvider = "numbers")
    public void testSquare(int input, int expected) {
        Assert.assertEquals(input * input, expected);
    }
}

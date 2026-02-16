import org.testng.annotations.Test;
import org.testng.annotations.DataProvider;

public class MethodRefTest {
    @DataProvider
    public Object[][] data() {
        return new Object[][] {{"hello"}, {"world"}};
    }

    @Test(dataProvider = "data")
    public void testStrings(String value) {
        assert value != null;
    }
}

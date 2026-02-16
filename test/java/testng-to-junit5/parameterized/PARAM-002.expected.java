import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.provider.MethodSource;

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

import org.testng.annotations.BeforeMethod;
import org.testng.annotations.AfterMethod;
import org.testng.annotations.Test;
import org.testng.Assert;

public class LifecycleTest {
    private String data;

    @BeforeMethod
    public void setUp() {
        data = "test";
    }

    @AfterMethod
    public void tearDown() {
        data = null;
    }

    @Test
    public void testData() {
        Assert.assertNotNull(data);
    }
}

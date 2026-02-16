import org.testng.annotations.*;
import org.testng.Assert;

public class CombinedTest {
    @BeforeMethod
    public void setUp() {
        System.out.println("setup");
    }

    @AfterMethod
    public void tearDown() {
        System.out.println("teardown");
    }

    @Test
    public void testBasic() {
        Assert.assertTrue(true);
    }

    @Test(enabled = false)
    public void testSkipped() {
        assert false;
    }
}

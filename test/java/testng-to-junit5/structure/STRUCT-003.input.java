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
    public void testAdd() {
        Assert.assertEquals(2, 1 + 1);
        Assert.assertTrue(true);
    }

    @Test
    public void testSubtract() {
        Assert.assertEquals(0, 1 - 1);
    }
}

import org.testng.annotations.Test;
import org.testng.Assert;

public class MultiMethodTest {
    @Test
    public void testFirst() {
        Assert.assertTrue(true);
    }

    @Test
    public void testSecond() {
        Assert.assertEquals(getResult(), 42);
    }

    @Test
    public void testThird() {
        Assert.assertNotNull(getObj());
    }
}

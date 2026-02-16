import org.testng.annotations.Test;
import org.testng.Assert;

public class MultiMethodTest {
    @Test
    public void testFirst() {
        Assert.assertEquals(1, 1);
    }

    @Test
    public void testSecond() {
        Assert.assertTrue(true);
    }
}

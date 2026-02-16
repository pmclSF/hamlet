import org.junit.Test;
import org.junit.Assert;

public class MultiMethodTest {
    @Test
    public void testFirst() {
        Assert.assertEquals(1, 1);
    }

    @Test
    public void testSecond() {
        Assert.assertEquals(2, 2);
    }

    @Test
    public void testThird() {
        Assert.assertNotNull("hello");
    }
}

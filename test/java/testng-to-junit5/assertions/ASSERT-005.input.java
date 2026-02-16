import org.testng.annotations.Test;
import org.testng.Assert;

public class MultiAssertTest {
    @Test
    public void testMultiple() {
        Assert.assertEquals(getResult(), 42);
        Assert.assertTrue(isValid());
        Assert.assertNotNull(getObj());
        Assert.assertFalse(isFalse());
    }
}

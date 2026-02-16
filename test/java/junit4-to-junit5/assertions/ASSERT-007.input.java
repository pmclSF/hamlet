import org.junit.Test;
import org.junit.Assert;

public class MultiAssertTest {
    @Test
    public void testMultiple() {
        Assert.assertEquals(1, getOne());
        Assert.assertTrue(isTrue());
        Assert.assertNotNull(getObj());
        Assert.assertFalse(isFalse());
    }
}

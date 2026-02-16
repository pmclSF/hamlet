import org.testng.annotations.Test;
import org.testng.Assert;

public class MultiAssertTest {
    @Test
    public void testMultiple() {
        Assert.assertNotNull(getResult());
        Assert.assertEquals(getResult().getValue(), 42);
        Assert.assertTrue(getResult().isValid());
    }
}

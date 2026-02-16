import org.junit.Test;
import org.junit.Assert;

public class AssertBooleanTest {
    @Test
    public void testBoolean() {
        Assert.assertTrue(isValid());
        Assert.assertFalse(isEmpty());
    }
}

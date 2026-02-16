import org.testng.annotations.Test;
import org.testng.Assert;

public class AssertNullTest {
    @Test
    public void testNull() {
        Assert.assertNull(getNullValue());
        Assert.assertNotNull(getObject());
    }
}

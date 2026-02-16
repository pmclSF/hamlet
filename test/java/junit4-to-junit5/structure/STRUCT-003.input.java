import org.junit.Before;
import org.junit.After;
import org.junit.Test;
import org.junit.Assert;

public class SetupTeardownTest {
    private String data;

    @Before
    public void setUp() {
        data = "test";
    }

    @After
    public void tearDown() {
        data = null;
    }

    @Test
    public void testData() {
        Assert.assertNotNull(data);
    }
}

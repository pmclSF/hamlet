import org.junit.After;
import org.junit.Test;

public class TeardownTest {

    private Object resource;

    @After
    public void tearDown() {
        resource = null;
    }

    @Test
    public void testResource() {
        resource = new Object();
        assertNotNull(resource);
    }
}

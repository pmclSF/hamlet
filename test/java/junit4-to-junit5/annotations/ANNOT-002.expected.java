import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Test;

public class TeardownTest {

    private Object resource;

    @AfterEach
    public void tearDown() {
        resource = null;
    }

    @Test
    public void testResource() {
        resource = new Object();
        assertNotNull(resource);
    }
}

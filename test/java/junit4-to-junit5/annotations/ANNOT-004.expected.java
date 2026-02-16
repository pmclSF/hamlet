import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.Test;

public class ClassTeardownTest {

    private static Object sharedResource;

    @AfterAll
    public static void tearDownClass() {
        sharedResource = null;
    }

    @Test
    public void testSharedResource() {
        sharedResource = new Object();
        assertNotNull(sharedResource);
    }
}

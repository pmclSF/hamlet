import org.junit.AfterClass;
import org.junit.Test;

public class ClassTeardownTest {

    private static Object sharedResource;

    @AfterClass
    public static void tearDownClass() {
        sharedResource = null;
    }

    @Test
    public void testSharedResource() {
        sharedResource = new Object();
        assertNotNull(sharedResource);
    }
}

import org.junit.BeforeClass;
import org.junit.Test;

public class ClassSetupTest {

    private static String sharedValue;

    @BeforeClass
    public static void setUpClass() {
        sharedValue = "shared";
    }

    @Test
    public void testSharedValue() {
        assertNotNull(sharedValue);
    }
}

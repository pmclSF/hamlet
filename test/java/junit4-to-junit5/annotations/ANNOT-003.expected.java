import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;

public class ClassSetupTest {

    private static String sharedValue;

    @BeforeAll
    public static void setUpClass() {
        sharedValue = "shared";
    }

    @Test
    public void testSharedValue() {
        assertNotNull(sharedValue);
    }
}

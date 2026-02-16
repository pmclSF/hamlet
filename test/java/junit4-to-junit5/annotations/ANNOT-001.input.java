import org.junit.Before;
import org.junit.Test;

public class SetupTest {

    private String value;

    @Before
    public void setUp() {
        value = "initialized";
    }

    @Test
    public void testValue() {
        assertNotNull(value);
    }
}

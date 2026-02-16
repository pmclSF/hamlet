import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

public class SetupTest {

    private String value;

    @BeforeEach
    public void setUp() {
        value = "initialized";
    }

    @Test
    public void testValue() {
        assertNotNull(value);
    }
}

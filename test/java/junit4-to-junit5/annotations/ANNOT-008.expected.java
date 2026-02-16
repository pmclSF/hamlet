import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Test;

public class CombinedAnnotationsTest {

    private String value;

    @BeforeEach
    public void setUp() {
        value = "ready";
    }

    @AfterEach
    public void tearDown() {
        value = null;
    }

    @Test
    public void testValue() {
        assertNotNull(value);
    }

    @Disabled
    @Test
    public void testSkipped() {
        fail("This test should be skipped");
    }
}

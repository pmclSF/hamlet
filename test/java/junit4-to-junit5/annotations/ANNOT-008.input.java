import org.junit.After;
import org.junit.Before;
import org.junit.Ignore;
import org.junit.Test;

public class CombinedAnnotationsTest {

    private String value;

    @Before
    public void setUp() {
        value = "ready";
    }

    @After
    public void tearDown() {
        value = null;
    }

    @Test
    public void testValue() {
        assertNotNull(value);
    }

    @Ignore
    @Test
    public void testSkipped() {
        fail("This test should be skipped");
    }
}

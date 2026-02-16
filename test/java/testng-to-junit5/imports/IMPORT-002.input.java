import org.testng.annotations.Test;
import static org.testng.Assert.*;

public class StaticImportTest {
    @Test
    public void testStatic() {
        assertEquals(42, 42);
    }
}

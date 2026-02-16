import org.testng.annotations.BeforeClass;
import org.testng.annotations.Test;

public class BeforeAllTest {
    @BeforeClass
    public static void setUpClass() {
    }

    @Test
    public void testSomething() {
        assert true;
    }
}

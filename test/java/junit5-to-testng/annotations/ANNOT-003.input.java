import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;

public class BeforeAllTest {
    @BeforeAll
    public static void setUpClass() {
    }

    @Test
    public void testSomething() {
        assert true;
    }
}

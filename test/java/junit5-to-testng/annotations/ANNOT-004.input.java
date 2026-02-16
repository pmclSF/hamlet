import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.Test;

public class AfterAllTest {
    @AfterAll
    public static void tearDownClass() {
    }

    @Test
    public void testSomething() {
        assert true;
    }
}

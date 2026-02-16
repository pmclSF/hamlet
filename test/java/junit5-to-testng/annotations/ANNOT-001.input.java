import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

public class BeforeTest {
    @BeforeEach
    public void setUp() {
        System.out.println("setup");
    }

    @Test
    public void testSomething() {
        assert true;
    }
}

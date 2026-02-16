import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Test;

public class AfterTest {
    @AfterEach
    public void tearDown() {
        System.out.println("teardown");
    }

    @Test
    public void testSomething() {
        assert true;
    }
}

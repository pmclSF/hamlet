import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.Test;

public class ClassTeardownTest {
    @AfterAll
    public static void tearDownClass() {
        System.out.println("class teardown");
    }

    @Test
    public void testSomething() {
        assert true;
    }
}

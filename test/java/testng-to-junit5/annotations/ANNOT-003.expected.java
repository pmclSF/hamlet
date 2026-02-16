import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;

public class ClassSetupTest {
    @BeforeAll
    public static void setUpClass() {
        System.out.println("class setup");
    }

    @Test
    public void testSomething() {
        assert true;
    }
}

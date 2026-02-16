import org.testng.annotations.BeforeMethod;
import org.testng.annotations.Test;

public class BeforeTest {
    @BeforeMethod
    public void setUp() {
        System.out.println("setup");
    }

    @Test
    public void testSomething() {
        assert true;
    }
}

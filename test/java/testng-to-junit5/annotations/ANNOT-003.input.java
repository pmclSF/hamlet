import org.testng.annotations.BeforeClass;
import org.testng.annotations.Test;

public class ClassSetupTest {
    @BeforeClass
    public static void setUpClass() {
        System.out.println("class setup");
    }

    @Test
    public void testSomething() {
        assert true;
    }
}

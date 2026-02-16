import org.testng.annotations.Factory;
import org.testng.annotations.Test;

public class FactoryTest {
    @Factory
    public Object[] createTests() {
        return new Object[] { new FactoryTest() };
    }

    @Test
    public void testSomething() {
        assert true;
    }
}

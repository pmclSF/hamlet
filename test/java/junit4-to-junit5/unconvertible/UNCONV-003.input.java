import org.junit.Test;
import org.junit.Rule;
import org.junit.rules.TestName;

public class TestNameRuleTest {
    @Rule
    public TestName testName = new TestName();

    @Test
    public void testGetName() {
        System.out.println(testName.getMethodName());
    }
}

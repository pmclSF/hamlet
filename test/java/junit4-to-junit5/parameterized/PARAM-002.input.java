import org.junit.Test;
import org.junit.Assert;
import org.junit.runner.RunWith;
import org.junit.runners.Parameterized;
import org.junit.runners.Parameterized.Parameters;

@RunWith(Parameterized.class)
public class ConstructorParamTest {
    private final int input;
    private final int expected;

    public ConstructorParamTest(int input, int expected) {
        this.input = input;
        this.expected = expected;
    }

    @Parameters
    public static Object[][] data() {
        return new Object[][] {{1, 1}, {2, 4}, {3, 9}};
    }

    @Test
    public void testSquare() {
        Assert.assertEquals(expected, input * input);
    }
}

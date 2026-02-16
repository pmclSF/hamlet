import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.runners.Parameterized;
import org.junit.runners.Parameterized.Parameters;

@ExtendWith(Parameterized.class)
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
        Assertions.assertEquals(expected, input * input);
    }
}

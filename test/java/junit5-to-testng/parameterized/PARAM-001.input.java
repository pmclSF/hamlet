import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.ValueSource;

public class ValueSourceTest {
    @ParameterizedTest
    @ValueSource(strings = {"hello", "world"})
    public void testStrings(String value) {
        Assertions.assertNotNull(value);
    }
}

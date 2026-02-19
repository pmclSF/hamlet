import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.NullAndEmptySource;

public class NullEmptyTest {
    @ParameterizedTest
    @NullAndEmptySource
    public void testNullEmpty(String value) {
        Assertions.assertTrue(value == null || value.isEmpty());
    }
}

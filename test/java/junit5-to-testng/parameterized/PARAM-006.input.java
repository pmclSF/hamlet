import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.ValueSource;
import org.junit.jupiter.params.provider.CsvSource;
import java.util.List;

public class MultiImportTest {
    @Test
    public void testBasic() {
        Assertions.assertTrue(true);
    }

    @ParameterizedTest
    @ValueSource(ints = {1, 2, 3})
    public void testInts(int value) {
        Assertions.assertTrue(value > 0);
    }
}

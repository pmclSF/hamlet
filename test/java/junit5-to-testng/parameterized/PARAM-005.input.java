import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.EnumSource;
import java.time.Month;

public class EnumSourceTest {
    @ParameterizedTest
    @EnumSource(Month.class)
    public void testMonth(Month month) {
        Assertions.assertNotNull(month);
    }
}

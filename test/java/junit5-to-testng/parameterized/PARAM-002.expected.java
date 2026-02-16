import org.testng.Assert;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.MethodSource;
import java.util.stream.Stream;

public class MethodSourceTest {
    static Stream<String> stringProvider() {
        return Stream.of("apple", "banana");
    }

    @ParameterizedTest
    @MethodSource("stringProvider")
    public void testFruit(String fruit) {
        Assert.assertNotNull(fruit);
    }
}

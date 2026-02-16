import org.junit.jupiter.api.Test;
import static org.hamcrest.CoreMatchers.is;
import static org.junit.jupiter.api.Assertions.assertThat;

public class HamcrestTest {
    @Test
    public void testHamcrest() {
        // HAMLET-TODO [UNCONVERTIBLE-HAMCREST]: Hamcrest assertThat with matchers is not directly convertible
        // Original: assertThat(getResult(), is(42));
        // Manual action required: Rewrite using JUnit 5 Assertions methods
        assertThat(getResult(), is(42));
    }
}

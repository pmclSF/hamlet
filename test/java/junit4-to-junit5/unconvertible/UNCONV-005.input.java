import org.junit.Test;
import static org.hamcrest.CoreMatchers.is;
import static org.junit.Assert.assertThat;

public class HamcrestTest {
    @Test
    public void testHamcrest() {
        assertThat(getResult(), is(42));
    }
}

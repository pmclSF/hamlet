import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;

public class SetupTeardownTest {
    private String data;

    @BeforeEach
    public void setUp() {
        data = "test";
    }

    @AfterEach
    public void tearDown() {
        data = null;
    }

    @Test
    public void testData() {
        Assertions.assertNotNull(data);
    }
}

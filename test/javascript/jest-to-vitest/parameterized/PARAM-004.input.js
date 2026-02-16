describe('Math operations', () => {
  it.each`
    a    | b    | expected
    ${1} | ${2} | ${3}
    ${3} | ${4} | ${7}
    ${5} | ${5} | ${10}
  `('add($a, $b) = $expected', ({ a, b, expected }) => {
    expect(a + b).toBe(expected);
  });

  it.each`
    input      | expected
    ${'hello'} | ${5}
    ${''}      | ${0}
    ${'test'}  | ${4}
  `('length of "$input" is $expected', ({ input, expected }) => {
    expect(input.length).toBe(expected);
  });
});

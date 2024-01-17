# Shortlink generation

[URL shortening algorithm](https://byby.dev/url-shortener)

1. MD5 hashing generates a 32-character string from the set `[a-fA-F0-9]`;

2. Base64 encode it to 32-character string from the set `[a-zA-z0-9+/=_]`;

3. Remove extra characters `[+/=_]`;

4. Trim resulting set to eight characters.

### Calculation probability of recurrence (*Pr*):

>$$ P_r = {1 \over A_n^k} = {(n−k)! \over n!} $$
>
>$$ n = 26 + 26 + 10 = 62 $$
>
>$$ k = 8 $$
>
>$$ P_r = 7,335363631522695e-15 $$

### The probability of recurrence is very very low.
This approach provides a good balance between uniqueness and simplicity.
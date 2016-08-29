def encrypt(plaintext, key, alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"):
	"""Encrypts and returns the plaintext encrypted with the key and alphabet given
	using the Vigenere cipher."""
	lk = len(key)
	la = len(alphabet)
	ciphertext = ""
	for pi in range(len(plaintext)):
		pj  = alphabet.index(plaintext[pi])
		ki = pi%lk
		kj = alphabet.index(key[ki])
		cj = (pj+kj)%la
		ciphertext += alphabet[cj]
	return ciphertext

def decrypt(ciphertext, key, alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"):
	"""Decrypts and returns the ciphertext decrypted with the key and alphabet given
	using the Vigenere cipher."""
	lk = len(key)
	la = len(alphabet)
	plaintext = ""
	for ci in range(len(ciphertext)):
		cj  = alphabet.index(ciphertext[ci])
		ki = ci%lk
		kj = alphabet.index(key[ki])
		pj = (cj-kj)%la
		plaintext += alphabet[pj]
	return plaintext

#if __name__ == "__main__":
#	plaintext = "attackatdawn"
#	key =       "lemonlemonle"
#	ciphertext = encrypt(plaintext, key)
#	deciphtext = decrypt(ciphertext, key)
#	print 'Our plaintext: "%s"\nOur ciphertext: "%s"\n(generated with the key "%s")\nOur deciphered text: "%s"'%(plaintext, ciphertext, key, deciphtext)

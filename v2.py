#-------------------------------------------------------------------------------
def decrypt(string,passkey):
    
    cNums = map(chr, range(65, 91));
    m = len(passkey);

    newString = "";

    for i in range(len(string)):
        iString = cNums.index(string[i]); 
        iPass =  cNums.index(passkey[i % m]);
        value = (iString - iPass) % len(cNums);
        newString += cNums[value];
    
    return newString;

#-------------------------------------------------------------------------------
def encrypt(string,passkey):
    
    cNums = map(chr, range(65, 91));
    m = len(passkey);

    newString = "";

    for i in range(len(string)):
        iString = cNums.index(string[i]); 
        iPass =  cNums.index(passkey[i % m]);
        value = (iString + iPass) % len(cNums);
        newString += cNums[value];
    
    return newString;
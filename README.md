# shafolder
Generates sha256 hash (or BIP39 words) for a file or folder  
For a folder, the output is the hash of the sorted hashes of the contents of files  
Filenames and foldernames do not contribute to the hash(es)  
BIP39 words are taken from Bitcoin Improvement Proposal 39  
BIP39 words are here used to provide a memorable mnemonic checksum of file contents  

## Usage:

shafolder.exe [-verbose] [-makecopy] [-o3de] [-bip39] MyFileOrFolder

## Options:

-verbose  
  In addition to the overall hash for a folder, show hashes of individual files  
  
-makecopy  
  Make a copy of the entire folder, with partial hashes prepended to each filename  
  
-o3de  
  Create a SHA256SUMS file containing the hashes of each file, suitable for making o3de packages  

-bip39  
  Instead of generating hex strings for hashes, encode using BIP39 words  

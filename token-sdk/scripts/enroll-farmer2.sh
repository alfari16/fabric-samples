fabric-ca-client register -u http://localhost:27054 --id.attrs full_name="EAI" \
 --id.name EAI --id.secret password --id.type client \
 --enrollment.type idemix --idemix.curve gurvy.Bn254
fabric-ca-client enroll -u http://EAI:password@localhost:27054  \
 -M "$(pwd)/keys/owner1/wallet/EAI/msp" --enrollment.type idemix \
 --idemix.curve gurvy.Bn254

fabric-ca-client register -u http://localhost:27054  --id.attrs full_name="PT Teknologi Untuk Pembudidaya" \
 --id.name TUP --id.secret password --id.type client \
 --enrollment.type idemix --idemix.curve gurvy.Bn254
fabric-ca-client enroll -u http://TUP:password@localhost:27054  \
 -M "$(pwd)/keys/owner2/wallet/TUP/msp" --enrollment.type idemix \
 --idemix.curve gurvy.Bn254

fabric-ca-client register -u http://localhost:27054 --id.attrs full_name="MUHAMMAD NURUL HUDA" \
 --id.name 511t7vh4bpdf --id.secret password --id.type client \
 --enrollment.type idemix --idemix.curve gurvy.Bn254
fabric-ca-client enroll -u http://511t7vh4bpdf:password@localhost:27054  \
 -M "$(pwd)/keys/owner1/wallet/511t7vh4bpdf/msp" --enrollment.type idemix \
 --idemix.curve gurvy.Bn254

# A new Secret Controller (WIP)

Currently working on building a small example controller to manage different kinds of secrets. Using Kubebuilder which implements controller-runtime, I have built out some of the functions needed to manage secrets. I am hoping to add the following functionality;

- Get the types of secret and for each, set the deadline for expiry and then rotate as necessary
- For generic, at functionality to patch the current secret with a newly generated secret, preserving the "key" it is referenced by
- For TLS secrets, when certificates expire create new certs and manage them accordingly
- For docker-registry secrets, much the same as point 2.

I'm working on this as and when I have some free time, but I aim to keep adding new features as I am learning Go!

Current state of affairs;

- [x] Managed to get some of the logic to take an individual secret which matches certain labels/namespaces/type
- [x] Generates an arbitrary value of a predetermined length of bytes (to be changed to getting the length of the current slice of bytes)
- [x] I realised that the complexity of a controller takes a lot of consideration of how you deal with different types of the given GVK and it's specs
- [x] This controller was a beginner one to familiarise myself with the libraries, and I am thinking of new controller examples to blog post on!

TODO;

- [ ] Publish some research on reversing one of the controllers in the core kubernetes repo
- [ ] Start working on another controller or rewrite the first one above to be better (without making assumption in the way I've coded it)

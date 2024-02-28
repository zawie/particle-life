# Particle Life

Particle life is a particle simulator implemented in Golang showcasing emergent behavior.

Each particle color is randomly assigned different repelling or attracting forces to other particles. Conservation of energy is not required and forces do not need to be equal (Newton is rolling in his grave!)

This simple system displays emergent behavior as life like structures form under these simple constraints.

https://github.com/zawie/particle-life/assets/15623191/30f744da-1bce-4f2e-9fd9-c57f416def51


## Optimizations

In theory, this requires `O(n^2)` computations per frame to compute the forces  where `n` is the number of particles. However, we can reduce this by chunking the plane into a grid. That way, for any given partlce we only need to scan grids that the particle is in or near. Additionally, for far enough chunks we memoize the "mass" of each color in the grid so we can further reduce the computation required per frame. This effectively reduces the time complexity to `O(n)`, assuming roughly uniform density. 

These optimizations enable the system to run thousands of particles operating at 60 frames per second

In the video below, the lines represent the actual comparisons the algorithm performs. observe how most pairs are never checked in any given frame.

https://github.com/zawie/particle-life/assets/15623191/e5617c82-fc54-4f4d-86b8-a377df327f97

An alternative route to chunking would be to implement a [QuadTree](https://en.wikipedia.org/wiki/Quadtree#:~:text=A%20quadtree%20is%20a%20tree,into%20four%20quadrants%20or%20regions.) data structure. Effectively, this is where we chunk the plane dynamically. This gaurentees as an `O(nlogn)` complexity, without needing to make any density assumptions. 

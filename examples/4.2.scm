(define chain-left (lambda (procedure right) (begin (define left (make-chan 0)) (go (procedure left right)) left)))
(define fac-i (lambda (left right) (begin (define n (<- left)) (if (equal? n 1) (-> left 1) (begin (-> right (- n 1)) (-> left (* n (<- right))))) (fac-i left right))))
(define make-facchain (lambda (i acc) (if (equal? i 0) acc (make-facchain (- i 1) (chain-left fac-i acc)))))
(define fac (lambda (x) (begin (define facport (make-facchain (+ x 1) 0)) (-> facport x) (<- facport))))
